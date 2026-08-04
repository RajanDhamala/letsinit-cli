from django.contrib import admin
from django.http import JsonResponse
from django.shortcuts import render
from django.urls import include, path


def read_root(request):
    return render(request, "index.html")


def health_check(request):
    return JsonResponse({"status": "ok"})


urlpatterns = [
    path("", read_root),
    path("admin/", admin.site.urls),
    path("health/", health_check),
    path("user/", include("apps.users.urls")),
]
