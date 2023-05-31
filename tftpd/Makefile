
NAME=tftpd
BUILD=go build -ldflags="-w -s"

all: ${NAME} docker-build docker-run docker-logs


${NAME}: main.go Makefile
	CGO_ENABLED=0 ${BUILD} -a -installsuffix cgo -o ${NAME} .
	
docker-build: Dockerfile ${NAME}
	docker build --no-cache -t ${NAME} .

docker-run: ${NAME}
#ifeq ($(shell docker images -q ${NAME} 2> /dev/null),)
	docker run -d --network=host --restart always --name ${NAME} -v /home/cat/projects/CISCO/tftp:/data ${NAME}:latest
#endif

docker-logs:
	docker logs -f ${NAME}

clean:	
	touch ${NAME}
	go clean
	docker rm -f ${NAME}
	docker rmi ${NAME}:latest
