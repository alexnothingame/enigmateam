import os
import httpx

CLIENT_ID = os.getenv("OAUTH_YANDEX_CLIENT_ID")
CLIENT_SECRET = os.getenv("OAUTH_YANDEX_CLIENT_SECRET")
REDIRECT_URI = os.getenv("NGROK_URL") + "/auth/yandex/callback"

def generate_yandex_oauth_redirect_uri(state: str):
    return (
        "https://oauth.yandex.ru/authorize"
        f"?response_type=code&client_id={CLIENT_ID}&redirect_uri={REDIRECT_URI}&state={state}"
    )

async def get_yandex_user_info(code: str):
    async with httpx.AsyncClient() as client:
        token_res = await client.post(
            "https://oauth.yandex.ru/token",
            data={
                "grant_type": "authorization_code",
                "code": code,
                "client_id": CLIENT_ID,
                "client_secret": CLIENT_SECRET,
            },
        )

        token = token_res.json().get("access_token")
        if not token:
            raise ValueError("Yandex token error")

        user_res = await client.get(
            "https://login.yandex.ru/info?format=json",
            headers={"Authorization": f"OAuth {token}"},
        )

        return user_res.json()
