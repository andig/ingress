############################
# STEP 1 build executable binary
############################
# golang alpine 1.11.5
FROM golang:alpine as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates tzdata alpine-sdk && update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

WORKDIR /build
COPY . .
COPY .git .git

# Fetch dependencies.

# Using go mod.
RUN go mod download
RUN make assets

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/ingress cmd/ingress/main.go

#############################
## STEP 2 build a small image
#############################
FROM scratch

# Import from builder.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# Copy our static executable
COPY --from=builder /go/bin/ingress /go/bin/ingress

# Use an unprivileged user.
USER appuser

# Run the binary.
ENTRYPOINT ["/go/bin/ingress"]
