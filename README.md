# Smart home controller

This is a repository for storing and processing smart-home sensors' events. We have different users and one-to many relationship between them and sensors. 
Each sensor has its type, current state, etc. Finally, we receive and processes the events (for instance, when the temperature has changed). 

For storing the stuff, we have in-memory and postgres databases.

On the last layer we have a server that provides rest [api](api/swagger.yaml) for getting and posting events, sensors and users.

Furthermore, all of that is packed in a docker container. 

# Build instructions
1. Build an app via `make controller-build`
2. Run database via `docker compose up -d`
3. Make migrations via `make migrate-up`
4. Run the app via `make controller-run`