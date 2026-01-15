from datetime import datetime, timedelta

auth_requests = {}

def create_auth_request(token_login: str):
    auth_requests[token_login] = {
        "expires_at": datetime.utcnow() + timedelta(minutes=5),
        "status": "pending",
        "access_token": None,
        "refresh_token": None,
    }

def set_auth_denied(token_login: str):
    if token_login in auth_requests:
        auth_requests[token_login]["status"] = "denied"

def set_auth_success(token_login: str, access: str, refresh: str):
    if token_login in auth_requests:
        auth_requests[token_login].update({
            "status": "success",
            "access_token": access,
            "refresh_token": refresh,
        })
