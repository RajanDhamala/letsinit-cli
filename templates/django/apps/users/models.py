import uuid

from django.db import models


class User(models.Model):
    ROLE_CHOICES = [
        ("user", "User"),
        ("admin", "Admin"),
        ("moderator", "Moderator"),
    ]

    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    username = models.CharField(max_length=150, unique=True)
    email = models.EmailField(max_length=50, unique=True)
    password = models.CharField(max_length=128)
    avatar = models.URLField(max_length=1000, blank=True, null=True)
    role = models.CharField(max_length=10, choices=ROLE_CHOICES, default="user")
    google_id = models.CharField(
        max_length=255,
        blank=True,
        null=True,
        unique=True,
    )
    github_id = models.CharField(
        max_length=255,
        blank=True,
        null=True,
        unique=True,
    )
    provider_id = models.CharField(max_length=255, blank=True, null=True)
    provider_name = models.CharField(max_length=50, blank=True, null=True)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    def __str__(self):
        return self.username
