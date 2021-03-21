# Authentication

Credits: https://gist.github.com/mschoebel/9398202

## Functionality
Login window for debugging
No checking credentials against data base, any username or password will do
A secure cookie will be created backend for session management
Passwords are sent in clear text right now, as long as it is done using https it should be ok according to internet.


## Endpoints
* ``/login`` - A login form
* ``/auth`` - Generates coockie, redirects to /home
* ``/home`` - A debug view that shows when a user is authenticated
* ``/logout`` - An endpoint to call to clear the cookie and log out

## Todo
Hash password with salt in backend
Compare password and user against database
Implement JWT token for auth