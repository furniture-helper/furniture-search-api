FROM golang:1.26 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap ./cmd/api

FROM public.ecr.aws/lambda/provided:al2023
COPY --from=build /app/bootstrap ${LAMBDA_RUNTIME_DIR}/bootstrap
CMD [ "bootstrap" ]