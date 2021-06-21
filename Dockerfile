### Build phase
FROM vshn-registry-mirror.appuioapp.ch/library/golang:1.16.5 AS build

# Will be printed in stacktraces
WORKDIR /openshift-machineset-egress-cidr-operator

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build application
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /operator .


### Runtime phase
FROM vshn-registry-mirror.appuioapp.ch/library/busybox AS runtime
WORKDIR /
CMD [ "/operator" ]
COPY --from=build /operator /operator
