from fastapi import APIRouter, HTTPException, Body
from fastapi.responses import HTMLResponse

from db import users
from auth_state import (
    create_auth_request,
    set_auth_denied,
    set_auth_success,
    auth_requests,
)
from oauth_github import generate_github_oauth_redirect_uri, get_github_user_info
from oauth_yandex import generate_yandex_oauth_redirect_uri, get_yandex_user_info
from jwt_utils import create_access_token, create_refresh_token, decode_token
from code_auth import generate_code, consume_code

router = APIRouter(
    prefix="/auth",
    tags=["Authentication"],
)

@router.post(
    "/oauth/start",
    summary="Start OAuth authentication",
    description="""

    Запускает процесс входа через выбранного провайдера (GitHub или Яндекс).Создаёт запрос 
    на авторизацию с идентификатором token_login и возвращает ссылку, 
    на которую нужно отправить пользователядля завершения входа.
    """,
)
def oauth_start(provider: str, token_login: str):
    create_auth_request(token_login)

    if provider == "github":
        url = generate_github_oauth_redirect_uri(token_login)
    elif provider == "yandex":
        url = generate_yandex_oauth_redirect_uri(token_login)
    else:
        raise HTTPException(400, "Unknown provider")

    return {"url": url}


@router.get(
    "/github/callback",
    summary="GitHub OAuth callback",
    description="""

    Обрабатывает ответ от GitHub после входа пользователя. Получает данные пользователя, 
    создаёт его в базе при необходимости, выдаёт JWT-токены и помечает вход как успешный для указанного token_login.
    """,
)
async def github_callback(
    code: str | None = None,
    state: str | None = None,
    error: str | None = None,
):
    if error:
        set_auth_denied(state)
        return HTMLResponse("GitHub authorization denied", 400)

    info = await get_github_user_info(code)
    email = info.get("email") or "no_email@github"

    user = await users.find_one({"email": email})
    if not user:
        count = await users.count_documents({})
        await users.insert_one({
            "email": email,
            "username": f"Аноним{count + 1}",
            "roles": ["student"],
            "refresh_tokens": [],
        })

    access_token = create_access_token(["read"])
    refresh_token = create_refresh_token(email)

    await users.update_one(
        {"email": email},
        {"$push": {"refresh_tokens": refresh_token}},
    )

    set_auth_success(state, access_token, refresh_token)

    return HTMLResponse("GitHub authorization successful")


@router.get(
    "/yandex/callback",
    summary="Yandex OAuth callback",
    description="""

    Обрабатывает ответ от Яндекса после входа пользователя. Получает информацию о пользователе, 
    создаёт или обновляет запись в базе, выдаёт JWT-токены и завершает процесс авторизации для token_login.
    """,
)
async def yandex_callback(
    code: str | None = None,
    state: str | None = None,
    error: str | None = None,
):
    if error:
        set_auth_denied(state)
        return HTMLResponse("Yandex authorization denied", 400)

    info = await get_yandex_user_info(code)
    email = info.get("default_email") or "no_email@yandex"

    user = await users.find_one({"email": email})
    if not user:
        count = await users.count_documents({})
        await users.insert_one({
            "email": email,
            "username": f"Аноним{count + 1}",
            "roles": ["student"],
            "refresh_tokens": [],
        })

    access_token = create_access_token(["read"])
    refresh_token = create_refresh_token(email)

    await users.update_one(
        {"email": email},
        {"$push": {"refresh_tokens": refresh_token}},
    )

    set_auth_success(state, access_token, refresh_token)

    return HTMLResponse("Yandex authorization successful")


@router.post(
    "/code/start",
    summary="Start code-based authentication",
    description="""

    Запускает авторизацию с помощью одноразового числового кода. Создаёт запрос 
    на вход и генерирует короткоживущий код, который пользователь должен подтвердить.
    """,
)
def code_start(token_login: str):
    create_auth_request(token_login)
    code = generate_code(token_login)
    return {"code": code}


@router.post(
    "/code/confirm",
    summary="Confirm code-based authentication",
    description="""
    
    Подтверждает вход с помощью одноразового кода и refresh-токена.Если 
    код и токен валидны, выдаёт новые JWT-токены и завершает авторизацию.
    """,
)
def code_confirm(code: str, refresh_token: str):
    token_login = consume_code(code)
    if not token_login:
        raise HTTPException(400, "Invalid or expired code")

    payload = decode_token(refresh_token)
    email = payload.get("email")
    if not email:
        raise HTTPException(400, "Invalid refresh token")

    access_token = create_access_token(["read"])
    refresh_token = create_refresh_token(email)

    set_auth_success(token_login, access_token, refresh_token)
    return {"status": "ok"}


@router.get(
    "/status",
    summary="Get authentication status",
    description="""

    Возвращает текущее состояние авторизации для token_login. Клиент может периодически 
    вызывать этот endpoint, чтобы узнать, завершён ли вход и получить JWT-токены после успеха.
    """,
)
def auth_status(token_login: str):
    entry = auth_requests.get(token_login)
    if not entry:
        raise HTTPException(404, "Unknown token_login")
    return entry


@router.post(
    "/refresh",
    summary="Refresh JWT tokens",
    description="""
    Обновляет access и refresh токены.
    Используется, если основной сервис вернул 401 Unauthorized.
    """
)
async def refresh_tokens(refresh_token: str = Body(..., embed=True)):

    payload = decode_token(refresh_token)
    if not payload:
        raise HTTPException(status_code=401, detail="Invalid refresh token")

    email = payload.get("email")
    if not email:
        raise HTTPException(status_code=401, detail="Invalid refresh token payload")

    user = await users.find_one({"email": email})
    if not user:
        raise HTTPException(status_code=404, detail="User not found")

    new_access_token = create_access_token(["read"])
    new_refresh_token = create_refresh_token(email)

    await users.update_one(
        {"email": email},
        {"$push": {"refresh_tokens": new_refresh_token}}
    )

    return {
        "access_token": new_access_token,
        "refresh_token": new_refresh_token
    }
