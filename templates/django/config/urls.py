from django.contrib import admin
from django.http import JsonResponse
from django.urls import path


def read_root(request):
    return JsonResponse({"message": "Django server is running"})


def health_check(request):
    return JsonResponse({"status": "ok"})


urlpatterns = [
    path("admin/", admin.site.urls),
    path("health/", health_check),
    path("", read_root),
]
