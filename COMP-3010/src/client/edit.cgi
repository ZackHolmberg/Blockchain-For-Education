#!/usr/bin/python

import sys
import os
from os import environ

request = sys.stdin
info = request.read()
output = "placeholder"
noLetters = False

if "&" in info:
    temp = info.split("&")
    temp = temp[0].split("=")[1].split("+")
    delimeter = ' '
    output = delimeter.join(temp) 
    if any(c.isalpha() for c in output):
        output = output + "|attending"
    else:
        noLetters = True

else:
    temp = info.split("=")[1].split("+")
    delimeter = ' '
    output = delimeter.join(temp)
    if any(c.isalpha() for c in output):
        output = output + "|not"
    else:
        noLetters = True

if noLetters:
    print("Content-Type: text/html\r\n\r\n")
    print("<html>")
    print("<head>")
    print("<title>COMP 3010 A1</title>")
    print("</head>")
    print("<body>")
    print("<h1><strong>Error:</strong> Name is invalid. Names must contain at least one letter. Please go back to the page and try again.</h1>")
    print("</body>")
    print("</html>")

else:
    cookies = os.environ["HTTP_COOKIE"].split(";")
    oldName = cookies[1].split("=")[1]
    oldStatus = cookies[2].split("=")[1]
    oldEntry = oldName + "|" + oldStatus
    # Read in whole file, and write back all except the one we want to replace
    # And then append new entry to fileand change cookie data
    with open("responses.txt", "r") as theFile:
        fileContents = theFile.readlines()
    with open("responses.txt", "w") as theFile:
        for entry in fileContents:
            if entry.strip("\n") != oldEntry:
                theFile.write(entry)
    file = open("responses.txt", "a")
    file.write(output+"\n")
    file.close()
    temp = output.split("|")
    name = temp[0]
    status = temp[1]
    print("Content-Type: text/html")
    print("Set-Cookie:Name = "+name+";")
    print("Set-Cookie:Status = "+status+";")
    print("Set-Cookie:LastResponse = "+output+";")
    print("Status: 303 See other")
    print("Location: https://www-test.cs.umanitoba.ca/~holmbezt/cgi-bin/index.cgi")
    print("\r")