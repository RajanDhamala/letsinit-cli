#!/usr/bin/env python
import os
import sys

from dotenv import load_dotenv

def main():
    load_dotenv()
    os.environ.setdefault("DJANGO_SETTINGS_MODULE", "config.settings")

    if len(sys.argv) == 2 and sys.argv[1] == "runserver":
        host = os.getenv("HOST", "0.0.0.0")
        port = os.getenv("PORT", "8000")
        sys.argv.append(f"{host}:{port}")

    from django.core.management import execute_from_command_line

    execute_from_command_line(sys.argv)


if __name__ == "__main__":
    main()
