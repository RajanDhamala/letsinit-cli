from django.urls import path

from . import views

urlpatterns = [
    path("", views.home),
    path("register/" ,views.register,name="register"),
    path("login/",views.login,name="login"),
    path("me/",views.AuthMe,name="me"),
    path("logout/",views.LogoutUser,name="logout"),
]
