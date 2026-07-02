import json

with open("mobile/package.json", "r+") as f:
    d = json.load(f)
    if "scripts" not in d:
        d["scripts"] = {}
    d["scripts"]["test"] = "jest"
    if "devDependencies" not in d:
        d["devDependencies"] = {}
    d["devDependencies"]["jest"] = "^29.7.0"
    d["devDependencies"]["jest-expo"] = "~52.0.3"
    d["devDependencies"]["@types/jest"] = "^29.5.12"
    f.seek(0)
    json.dump(d, f, indent=2)
    f.truncate()
