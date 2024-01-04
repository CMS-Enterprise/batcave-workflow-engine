# Use the official Nginx image as a base
FROM nginx:latest

# Remove default Nginx configuration file
RUN rm /etc/nginx/conf.d/default.conf

# Copy a new configuration file from your conf
COPY nginx.conf /etc/nginx/conf.d

# Copy your HTML file to the Nginx server
COPY index.html /usr/share/nginx/html/index.html

# Expose port 80
EXPOSE 80

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD curl -f http://localhost/health || exit 1

# Start Nginx when the container has provisioned.
CMD ["nginx", "-g", "daemon off;"]
