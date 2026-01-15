from fastapi import APIRouter, HTTPException, Body
from fastapi.responses import HTMLResponse
from datetime import datetime
from bson import ObjectId
from bson.errors import InvalidId
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

@router.get("/github/callback")
async def github_callback(code: str, state: str, error: str | None = None):
    if error:
        set_auth_denied(state)
        return HTMLResponse("GitHub authorization denied", status_code=400)

    info = await get_github_user_info(code)
    email = info.get("email") or "no_email@github"

    user = await users.find_one({"email": email})

    if not user:
        res = await users.insert_one({
            "email": email,
            "username": info.get("login", "Аноним"),
            "roles": ["student"],
            "refresh_token": None,
        })
        user_id = res.inserted_id
    else:
        user_id = user["_id"]

    access_token = create_access_token(str(user_id), ["read"])
    refresh_token = create_refresh_token(str(user_id))

    await users.update_one(
        {"_id": user_id},
        {"$set": {"refresh_token": refresh_token}},
    )

    set_auth_success(state, access_token, refresh_token)
    return HTMLResponse("GitHub authorization successful")


@router.get("/yandex/callback")
async def yandex_callback(code: str, state: str, error: str | None = None):
    if error:
        set_auth_denied(state)
        return HTMLResponse("Yandex authorization denied", status_code=400)

    info = await get_yandex_user_info(code)
    email = info.get("default_email")

    if not email:
        set_auth_denied(state)
        return HTMLResponse("Yandex did not return email", status_code=400)

    user = await users.find_one({"email": email})

    if not user:
        res = await users.insert_one({
            "email": email,
            "username": info.get("login", "Аноним"),
            "roles": ["student"],
            "refresh_token": None,
        })
        user_id = res.inserted_id
    else:
        user_id = user["_id"]

    access_token = create_access_token(str(user_id), ["read"])
    refresh_token = create_refresh_token(str(user_id))

    await users.update_one(
        {"_id": user_id},
        {"$set": {"refresh_token": refresh_token}},
    )

    set_auth_success(state, access_token, refresh_token)
    return HTMLResponse("Yandex authorization successful")


@router.post("/code/start")
def code_start(token_login: str):
    create_auth_request(token_login)
    code = generate_code(token_login)
    return {"code": code}


@router.post("/code/confirm")
async def code_confirm(code: str, refresh_token: str):
    token_login = consume_code(code)
    if not token_login:
        raise HTTPException(400, "Invalid or expired code")

    payload = decode_token(refresh_token)
    if not payload or payload.get("type") != "refresh":
        raise HTTPException(400, "Invalid refresh token")

    user_id = payload.get("sub")
    if not user_id:
        raise HTTPException(400, "Invalid refresh token")

    user = await users.find_one({
        "_id": ObjectId(user_id),
        "refresh_token": refresh_token,
    })

    if not user:
        raise HTTPException(401, "Refresh token revoked")

    access_token = create_access_token(user_id, ["read"])
    new_refresh = create_refresh_token(user_id)

    await users.update_one(
        {"_id": ObjectId(user_id)},
        {"$set": {"refresh_token": new_refresh}},
    )

    set_auth_success(token_login, access_token, new_refresh)
    return {"status": "ok"}


@router.get("/status")
def auth_status(token_login: str):
    entry = auth_requests.get(token_login)

    if not entry:
        return {"status": "expired"}

    if entry["status"] == "success":
        return entry

    if entry["expires_at"] < datetime.utcnow():
        auth_requests.pop(token_login, None)
        return {"status": "expired"}

    return {"status": "pending"}

@router.post("/refresh")
async def refresh(refresh_token: str = Body(..., embed=True)):
    payload = decode_token(refresh_token)

    if not payload or payload.get("type") != "refresh":
        raise HTTPException(401, "Invalid refresh token")

    user_id = payload.get("sub")

    try:
        user_obj_id = ObjectId(user_id)
    except InvalidId:
        raise HTTPException(401, "Invalid refresh token")

    user = await users.find_one({
        "_id": user_obj_id,
        "refresh_token": refresh_token,
    })

    if not user:
        raise HTTPException(401, "Refresh token revoked")

    new_access = create_access_token(user_id, ["read"])
    new_refresh = create_refresh_token(user_id)

    await users.update_one(
        {"_id": user_obj_id},
        {"$set": {"refresh_token": new_refresh}},
    )

    return {
        "access_token": new_access,
        "refresh_token": new_refresh,
    }
