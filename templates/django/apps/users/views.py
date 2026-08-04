from django.http import JsonResponse
from django.views.decorators.http import require_GET
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_POST
from django.contrib.auth.hashers import make_password
from django.contrib.auth.hashers import check_password
from django.http import HttpResponse
from utils.jwt_utils import (
    create_access_token,
    create_refresh_token,
)
from utils.auth import require_auth

import json

from django.http import JsonResponse
from .models import User

@require_GET
def home(request):
    return JsonResponse({"message": "User endpoints are up and running"})

@csrf_exempt
@require_POST
def register(request):
    body = json.loads(request.body)

    username = body.get("username")
    email = body.get("email")
    password = body.get("password")

    hashed_password = make_password(password)

    print("username:", username)
    print("email:", email)
    

    try:
        User.objects.get(email=email)
        return JsonResponse({"message":"invalid email or already exists"})
    except User.DoesNotExist:

        newUser=User.objects.create(
        username=username,
        email=email,
        password=hashed_password
         )

    return JsonResponse({
      "message": "User created successfully",
      "user": {
        "id": str(newUser.id),
        "username": newUser.username,
        "email": newUser.email,
        "role": newUser.role,
        "avatar": newUser.avatar,
      },
})

@csrf_exempt
@require_POST

def login (requset):
    body=json.loads(requset.body)
    email=body.get("email")

    if not email:
        return JsonResponse({"message":"Email is required"},status=400)
    # no regex checks rn
    plainpassword=body.get("password")
    try :
        existingUser=User.objects.get(email=email)
        isRight=check_password(plainpassword,existingUser.password)

        if not isRight:
            return JsonResponse({"message":"invalid credentials"},status=400)

        payload={
                "username":existingUser.username,
                "email":existingUser.email,
                "id":str(existingUser.id),
                "avatar":existingUser.avatar
                }
        newAccesstoken=create_access_token(payload)
        newRefreshtoken=create_refresh_token(payload)
        response=JsonResponse({"message":"login successful"})
        response.set_cookie(
                key="access_token",
                value=newAccesstoken,
                httponly=True,
                secure=False,
                max_age=15*60)

        response.set_cookie(
                key="refresh_token",
                value=newRefreshtoken,
                httponly=True,
                secure=False,
                max_age=60*60*24*7
                )
        return response
    except User.DoesNotExist:
        return JsonResponse({"message":"invalid credentials"},status=404)


@csrf_exempt
@require_GET
@require_auth
def AuthMe(request):
    user = request.auth
    return JsonResponse({
        "message": "user found in session",
        "data": {
            "id": user["id"],
            "username": user["username"],
            "email": user["email"],
            "avatar": user.get("avatar"),
        }
    })


@require_GET
@require_auth
def LogoutUser (request):
    response=JsonResponse({"message":"user logged out successfully"})
    response.delete_cookie("refresh_token")
    response.delete_cookie("access_token")
    return response


