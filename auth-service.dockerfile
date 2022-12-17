FROM alpine:latest
RUN mkdir /app

COPY ./auth-service/auth-service-bin /app

# Run the server executable
CMD [ "/app/auth-service-bin" ]