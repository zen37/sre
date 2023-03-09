# Welcome to your Bootcamp SRE Capstone Project!

Remember that you can find the complete instructions for this project **[here](https://classroom.google.com/w/MzgwNTc4MDgwMjAw/t/all)**.

If you have any questions, feel free to contact your mentor or one of us: Juan Barbosa, Laura Mata, or Francisco Bueno. We are here to support you.

## curl
curl -X POST -H "Content-Type: application/json" -d '{"username": "admin", "password": "secret"}' http://localhost:8000/login

export TOKEN=$(curl -X POST -H "Content-Type: application/json" -d '{"username": "admin", "password": "secret"}' http://localhost:8000/login | jq -r '.token')
curl -H "Authorization: Bearer $TOKEN" -H "Accept: application/json" "http://localhost:8000/mask-to-cidr?value=255.255.0.0"

export TOKEN=$(curl -X POST -H "Content-Type: application/json" -d '{"username": "admin", "password": "secret"}' http://localhost:8000/login | jq -r '.token')
curl -H "Authorization: Bearer $TOKEN" -H "Accept: application/json" "http://localhost:8000/cidr-to-mask?value=24"