FROM python:3.10-slim

WORKDIR /app

# Install Poetry
RUN pip install poetry

# Copy only requirements to cache them in docker layer
COPY pyproject.toml poetry.lock* ./

# Configure poetry to not use a virtual environment
RUN poetry config virtualenvs.create false

# Install dependencies without installing the root project
RUN poetry install --without dev --no-root

# Copy project
COPY . .

# Now install the project
RUN poetry install --without dev

# Create volume for persistent data
VOLUME /app/data

# Set environment variables
ENV PYTHONUNBUFFERED=1

# Run the bot
CMD ["python", "-m", "irlcord.main"] 