#!/usr/bin/python

import sys
import os
from os import environ

if not "HTTP_COOKIE" in os.environ:
    # no form submission so display form
    print("Content-type:text/html\r\n\r\n")
    print("<html>")
    print("<head>")
    print("<title>Ledger</title>")
    print("</head>")
    print("<body>")
    print("<h1>Add a transaction</h1>")
    # TODO: Change form action to localhost
    print("<form action=\"https://www-test.cs.umanitoba.ca/~holmbezt/cgi-bin/redirect.cgi\" method=\"POST\">")

    print("<label for=\"To\">To:</label><br>")
    print("<input type=\"text\" id=\"to\" name=\"to\"><br>")

    print("<label for=\"From\">From:</label><br>")
    print("<input type=\"text\" id=\"from\" name=\"from\"><br>")

    print("<label for=\"Amount\">Amount:</label><br>")
    print("<input type=\"text\" id=\"amount\" name=\"amount\"><br>")

    print("<input type=\"submit\" value=\"Submit\">")
    print("</form>")
    print("</body>")
    print("</html>")
else:
    # form has been submitted, so we have a cookie, but must check if user has anonymized or not
    cookies = os.environ["HTTP_COOKIE"].split(";")
    # If there is only one cookie, then the user has randomized, so display the form with previous response
    if len(cookies) == 1:
        lastResponse = cookies[0].split("=")[1].split("|")

        toVal = lastResponse[0]
        fromVal = lastResponse[1]
        amountVal = lastResponse[2]

        print("Content-type:text/html\r\n\r\n")
        print("<html>")
        print("<head>")
        print("<title>Ledger</title>")
        print("</head>")
        print("<body>")
        print("<h1>Add a transaction</h1>")
        print("<form action=\"https://www-test.cs.umanitoba.ca/~holmbezt/cgi-bin/redirect.cgi\" method=\"POST\">")

        print("<label for=\"To\">To:</label><br>")
        print("<input type=\"text\" id=\"to\" name=\"to\"><br>")

        print("<label for=\"From\">From:</label><br>")
        print("<input type=\"text\" id=\"from\" name=\"from\"><br>")

        print("<label for=\"Amount\">Amount:</label><br>")
        print("<input type=\"text\" id=\"amount\" name=\"amount\"><br>")

        print("<input type=\"submit\" value=\"Submit\">")
        print("</form>")

        print("<h4>Most recently added transaction: To - " +
              toVal+", From - "+fromVal+", Amount - "+amountVal+"</h4>")
        print("</body>")
        print("</html>")

    else:
        name = cookies[0].split("=")[1]
        status = cookies[1].split("=")[1]
        if status == "not":
            status = "not attending"
        # get data froom cookie and display
        print("Content-type:text/html\r\n\r\n")
        print("<html>")
        print("<head>")
        print("<title>COMP 3010 A1</title>")
        print("</head>")
        print("<body>")
        print("<h1>Thanks for your response!</h1>")
        print("<br><h2>You responded with:</h2>")
        print("<h3>Name: "+name+"</h3>")
        print("<h3>Status: "+status+"</h3>")
        print("</body>")
        print("</html>")
