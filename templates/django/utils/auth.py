# utils/auth.py

from functools import wraps

import jwt
from django.http import JsonResponse

from utils.jwt_utils import (
    create_access_token,
    validate_access_token,
    validate_refresh_token,
)


def require_auth(view_function):
    @wraps(view_function)
    def wrapper(request, *args, **kwargs):
        access_token = request.COOKIES.get("access_token")
        refresh_token = request.COOKIES.get("refresh_token")

        payload = None
        new_access_token = None

        if access_token:
            try:
                payload = validate_access_token(access_token)

            except jwt.ExpiredSignatureError:
                # Access token expired, continue to refresh-token check
                pass

            except jwt.InvalidTokenError:
                return JsonResponse(
                    {"message": "Invalid access token"},
                    status=401,
                )

        if payload is None:
            if not refresh_token:
                return JsonResponse(
                    {"message": "Authentication required"},
                    status=401,
                )

            try:
                refresh_payload = validate_refresh_token(refresh_token)

                payload = {
                    "id": refresh_payload["id"],
                    "username": refresh_payload["username"],
                    "email": refresh_payload["email"],
                }

                new_access_token = create_access_token(payload)

            except jwt.ExpiredSignatureError:
                return JsonResponse(
                    {"message": "Session expired. Please log in again."},
                    status=401,
                )

            except jwt.InvalidTokenError:
                return JsonResponse(
                    {"message": "Invalid refresh token"},
                    status=401,
                )

        request.auth = payload

        response = view_function(request, *args, **kwargs)

        if new_access_token:
            response.set_cookie(
                key="access_token",
                value=new_access_token,
                httponly=True,
                secure=False,  # True in production
                max_age=15 * 60,
            )

        return response

    return wrapper
