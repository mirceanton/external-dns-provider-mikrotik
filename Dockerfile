FROM python:3.12-slim

ENV PYTHONUNBUFFERED true

# Copy the requirements file and pip install it
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Set the working directory in the container
WORKDIR /app

# Copy over the source code files
COPY src/ .

# Make port 8088 available to the world outside this container
EXPOSE 8088

# Run main.py when the container launches
ENTRYPOINT ["python3", "main.py"]
