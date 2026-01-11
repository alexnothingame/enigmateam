import random
from datetime import datetime, timedelta

codes = {}

def generate_code(token_login: str):
    code = str(random.randint(10000, 99999))
    codes[code] = {
        "token_login": token_login,
        "expires_at": datetime.utcnow() + timedelta(minutes=1)
    }
    return code

def consume_code(code: str):
    entry = codes.get(code)
    if not entry or entry["expires_at"] < datetime.utcnow():
        return None
    return entry["token_login"]

