from django.contrib import admin

from apps.users.models import User


@admin.register(User)
class UserAdmin(admin.ModelAdmin):
    list_display = (
        "id",
        "username",
        "email",
        "provider_name",
        "created_at",
    )
    search_fields = ("username", "email", "provider_id")
    list_filter = ("provider_name",)
    readonly_fields = ("id", "created_at", "updated_at")
