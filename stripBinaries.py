from os import listdir, environ
from sys import argv

username = environ["username"].encode()
replacement = b"user"

for filename in listdir("bin"):
    if not filename.split(".")[0] == argv[1]: continue
    
    with open("bin/"+filename, "rb") as f:
        data = f.read().replace(username, replacement)
    with open("bin/"+filename, "wb") as f:
        f.write(data)
