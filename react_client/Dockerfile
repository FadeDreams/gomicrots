FROM node:21-alpine
# Set working directory
WORKDIR /app

# Copy package.json and install dependencies using pnpm
COPY package.json .
RUN npm install -g pnpm && \
    pnpm install --only=prod

# Copy the rest of the application code
COPY . .

EXPOSE 3000

# Specify the command to run your application
CMD ["pnpm", "run", "start"]
