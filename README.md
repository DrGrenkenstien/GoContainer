## Steps to Run the container:
 - Adfter copying the project locally, run the command: sudo ./bin/containerproject create {container_id : string} /bin/bash to create a container with hostname container_id
 - Then Run the command: sudo ./bin/containerproject run {container_id} to run already created container. This step will show no container {container_id} if a container with this id doesn't exist
 - After running the container, user can execute any bin/bash command
 - Do ctrl + D to exit from the container
