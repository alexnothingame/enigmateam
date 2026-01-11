from fastapi import APIRouter, HTTPException
from fastapi.responses import HTMLResponse

from db import users
from auth_state import create_auth_request, set_auth_denied, set_auth_success, auth_requests
from oauth_github import generate_github_oauth_redirect_uri, get_github_user_info
from oauth_yandex import generate_yandex_oauth_redirect_uri, get_yandex_user_info
from jwt_utils import create_access_token, create_refresh_token, decode_token
from code_auth import generate_code, consume_code

router = APIRouter(prefix="/auth", tags=["Authentication"])

@router.post("/oauth/start")
def oauth_start(provider: str, token_login: str):
    create_auth_request(token_login)

    if provider == "github":
        url = generate_github_oauth_redirect_uri(token_login)
    elif provider == "yandex":
        url = generate_yandex_oauth_redirect_uri(token_login)
    else:
        raise HTTPException(400, "Unknown provider")

    return {"url": url}


async def _issue_tokens(email: str):
    user = await users.find_one({"email": email})
    if not user:
        res = await users.insert_one({
            "email": email,
            "roles": ["student"],
            "permissions": ["read"],
            "refresh_tokens": [],
        })
        user_id = str(res.inserted_id)
    else:
        user_id = str(user["_id"])

    access = create_access_token(user_id, ["read"])
    refresh = create_refresh_token(user_id, email)

    await users.update_one(
        {"_id": user["_id"]},
        {"$push": {"refresh_tokens": refresh}},
    )

    return access, refresh


@router.get("/github/callback")
async def github_callback(code: str | None = None, state: str | None = None, error: str | None = None):
    if error:
        set_auth_denied(state)
        return HTMLResponse("GitHub denied", 400)

    info = await get_github_user_info(code)
    email = info.get("email") or "no_email@github"

    access, refresh = await _issue_tokens(email)
    set_auth_success(state, access, refresh)

    return HTMLResponse("GitHub auth success")


@router.get("/yandex/callback")
async def yandex_callback(code: str | None = None, state: str | None = None, error: str | None = None):
    if error:
        set_auth_denied(state)
        return HTMLResponse("Yandex denied", 400)

    info = await get_yandex_user_info(code)
    email = info.get("default_email") or "no_email@yandex"

    access, refresh = await _issue_tokens(email)
    set_auth_success(state, access, refresh)

    return HTMLResponse("Yandex auth success")


@router.post("/code/start")
def code_start(token_login: str):
    create_auth_request(token_login)
    return {"code": generate_code(token_login)}


@router.post("/code/confirm")
def code_confirm(code: str, refresh_token: str):
    token_login = consume_code(code)
    if not token_login:
        raise HTTPException(400, "Invalid code")

    payload = decode_token(refresh_token)
    if payload.get("type") != "refresh":
        raise HTTPException(400, "Invalid refresh token")

    access = create_access_token(payload["user_id"], ["read"])
    refresh = create_refresh_token(payload["user_id"], payload["email"])

    set_auth_success(token_login, access, refresh)
    return {"status": "ok"}


@router.get("/status")
def auth_status(token_login: str):
    entry = auth_requests.get(token_login)
    if not entry:
        raise HTTPException(404, "Unknown token_login")
    return entry

