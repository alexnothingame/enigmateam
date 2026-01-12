import os
import httpx

CLIENT_ID = os.getenv("OAUTH_GITHUB_CLIENT_ID")
CLIENT_SECRET = os.getenv("OAUTH_GITHUB_CLIENT_SECRET")
REDIRECT_URI = os.getenv("NGROK_URL") + "/auth/github/callback"

def generate_github_oauth_redirect_uri(state: str):
    return (
        "https://github.com/login/oauth/authorize"
        f"?client_id={CLIENT_ID}&redirect_uri={REDIRECT_URI}&state={state}"
    )

async def get_github_user_info(code: str):
    async with httpx.AsyncClient() as client:
        token_res = await client.post(
            "https://github.com/login/oauth/access_token",
            headers={"Accept": "application/json"},
            data={
                "client_id": CLIENT_ID,
                "client_secret": CLIENT_SECRET,
                "code": code,
                "redirect_uri": REDIRECT_URI,
            },
        )

        token = token_res.json()["access_token"]

        user_res = await client.get(
            "https://api.github.com/user",
            headers={"Authorization": f"Bearer {token}"},
        )

        return user_res.json()