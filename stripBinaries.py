from os import listdir, environ

username = environ["username"].encode()
replacement = b"user"

for filename in listdir("bin"):
    if not filename.endswith(".exe"): continue
    
    with open("bin/"+filename, "rb") as f:
        data = f.read().replace(username, replacement)
    with open("bin/"+filename, "wb") as f:
        f.write(data)
