
curl -X POST http://localhost:8080/upload \
-F "audio=@{audioFilepath};type=audio/mpeg" \
-F "yaml=@{yamlFilepath};type=application/x-yaml" \
-H "Accept: application/json"

curl -X POST http://localhost:8080/upload \
-F "audio=@/Users/gary/FCBH2024/download/ENGWEB/ENGWEBN2DA/B23___01_1John_______ENGWEBN2DA.mp3;type=audio/mpeg" \
-F "yaml=@/Users/gary/FCBH2024/request_post.yaml;type=application/x-yaml" \
-H "Accept: application/json"


