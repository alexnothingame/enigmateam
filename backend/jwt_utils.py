from datetime import datetime, timedelta
from jose import jwt

SECRET = "SUPER_SECRET_KEY"
ALGO = "HS256"

def create_access_token(user_id: int, permissions: list):
    return jwt.encode(
        {
            "user_id": user_id,
            "permissions": permissions,
            "exp": datetime.utcnow() + timedelta(minutes=5),
        },
        SECRET,
        ALGO,
    )


def create_refresh_token(email: str):
    return jwt.encode(
        {
            "email": email,
            "exp": datetime.utcnow() + timedelta(days=7),
        },
        SECRET,
        ALGO,
    )

def decode_token(token: str):
    try:
        return jwt.decode(token, SECRET, ALGO)
    except Exception:
        return None
