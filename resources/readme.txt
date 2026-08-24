HTTP Server Resources Directory
================================

This directory contains static files served by the HTTP server.

Directory Structure:
- *.html files: Served with text/html content-type (rendered in browser)
- *.txt files: Served as binary downloads (application/octet-stream)
- *.png, *.jpg files: Served as binary downloads
- uploads/: Directory for POST request file uploads

Adding Your Own Files:
1. HTML files will be rendered in the browser
2. Images and text files will trigger downloads
3. Unsupported file types will return 415 Unsupported Media Type

Security:
- All paths are validated to prevent directory traversal
- Only files in this directory and subdirectories can be accessed
- Directory listing is disabled for security

Testing:
- GET /: Returns index.html
- GET /about.html: Returns about page
- GET /sample.txt: Downloads this file
- POST /upload: Accepts JSON data and saves to uploads/

