from datetime import datetime, timedelta
from jose import jwt, JWTError

SECRET = "SUPER_SECRET_KEY"
ALGO = "HS256"


def create_access_token(user_id: str, permissions: list[str]):
    payload = {
        "sub": user_id,
        "type": "access",
        "permissions": permissions,
        "exp": datetime.utcnow() + timedelta(minutes=1),
    }
    return jwt.encode(payload, SECRET, ALGO)


def create_refresh_token(user_id: str):
    payload = {
        "sub": user_id,
        "type": "refresh",
        "exp": datetime.utcnow() + timedelta(days=7),
    }
    return jwt.encode(payload, SECRET, ALGO)


def decode_token(token: str):
    try:
        return jwt.decode(token, SECRET, algorithms=[ALGO])
    except JWTError:
        return None
