
from datetime import UTC, datetime, timedelta
from typing import Any

import jwt
from django.conf import settings


ALGORITHM = "HS256"


def create_token(
    payload: dict[str, Any],
    token_type: str,
    expires_in: timedelta,
) -> str:
    now = datetime.now(UTC)

    data = payload.copy()
    data.update({
        "type": token_type,
        "iat": now,
        "exp": now + expires_in,
    })

    return jwt.encode(
        data,
        settings.SECRET_KEY,
        algorithm=ALGORITHM,
    )


def create_access_token(payload: dict[str, Any]) -> str:
    return create_token(
        payload=payload,
        token_type="access",
        expires_in=timedelta(minutes=15),
    )


def create_refresh_token(payload: dict[str, Any]) -> str:
    return create_token(
        payload=payload,
        token_type="refresh",
        expires_in=timedelta(days=7),
    )


def validate_token(
    token: str,
    expected_type: str,
) -> dict[str, Any]:
    payload = jwt.decode(
        token,
        settings.SECRET_KEY,
        algorithms=[ALGORITHM],
    )

    if payload.get("type") != expected_type:
        raise jwt.InvalidTokenError("Wrong token type")

    return payload


def validate_access_token(token: str) -> dict[str, Any]:
    return validate_token(token, "access")


def validate_refresh_token(token: str) -> dict[str, Any]:
    return validate_token(token, "refresh")
