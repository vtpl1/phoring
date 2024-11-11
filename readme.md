ldd -v /pwd/my-app
ldd -v ./backend
nm -D ./backend | grep pthread_attr_getstacksize