FROM alpine:3.20.0
RUN mkdir /app
RUN mkdir /app/logs

COPY ./auth-service.bin /app

# Run the server executable
CMD [ "/app/auth-service.bin" ]