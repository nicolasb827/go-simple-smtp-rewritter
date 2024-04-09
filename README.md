This project is a tool to rewrite some parts of an email.
It takes "(Authenticated sender: <email>)" from Received header added by postfix to set a new From header.
The original From header is keep into X-Sender-From header.
It also change "+" addressing by removing "+" part into X-Enveloppe-To.

This tool was created for mail archive processing, not for production.

go-simple-smtp-rewritter by Nicolas B. is marked with CC0 1.0 