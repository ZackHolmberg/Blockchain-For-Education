#!/usr/bin/python

import sys
import os
from os import environ

print("Content-Type: text/html")
print("Set-Cookie:Name = ;Expires =Wed, 1 Jul 2020 07:28:00 GMT;")
print("Set-Cookie:Status = ; Expires =Wed, 1 Jul 2020 07:28:00 GMT;")
print("Status: 303 See other")
print("Location: https://www-test.cs.umanitoba.ca/~holmbezt/cgi-bin/index.cgi")
print("\r")