#!/bin/bash

# scaffold-bc.sh - Create a bounded context structure for DDD + Hexagonal Architecture
# Usage: bash scaffold-bc.sh <bounded-context-name>

set -e

# Check if bounded context name is provided
if [ -z "$1" ]; then
    echo "Error: Bounded context name is required"
    echo "Usage: bash scaffold-bc.sh <bounded-context-name>"
    echo "Example: bash scaffold-bc.sh user-management"
    exit 1
fi

BC_NAME="$1"
BASE_DIR="internal/${BC_NAME}"

# Check if the directory already exists
if [ -d "$BASE_DIR" ]; then
    echo "Error: Directory ${BASE_DIR} already exists"
    exit 1
fi

# Function to convert kebab-case to Title Case
to_title_case() {
    local input="$1"
    # Replace hyphens with spaces, then capitalize each word
    echo "$input" | sed -E 's/-/ /g' | sed 's/\b./\U&/g'
}

# Function to generate minimal agent.md
generate_agent_md() {
    local bc_name="$1"
    local bc_dir="$2"
    local title=$(to_title_case "$bc_name")

    cat > "${bc_dir}/agent.md" << EOF
# ${title}

**Type:**
**Description:**

**Status:**

---

## Ubiquitous Language

### **Main Entity**

---

## User Journeys

---

## Architecture

### Folder Structure

---

## Domain

### Aggregates

### Domain Events

### Repository Ports

---

## Application

### Commands

### Queries

---

## Adapters

### Driven (Secondary)

### Driving (Primary)

---

## Dependencies

### Consumes Events From

### Emits Events To

---

EOF
}

echo "Creating bounded context structure for: ${BC_NAME}"

# Create directory structure
mkdir -p "${BASE_DIR}/domain"
mkdir -p "${BASE_DIR}/application/command"
mkdir -p "${BASE_DIR}/application/query"
mkdir -p "${BASE_DIR}/adapter/driven/inmemory"
mkdir -p "${BASE_DIR}/adapter/driving/http"
mkdir -p "${BASE_DIR}/adapter/driving/consumer"

# Create .gitkeep files for empty directories
touch "${BASE_DIR}/application/command/.gitkeep"
touch "${BASE_DIR}/application/query/.gitkeep"
touch "${BASE_DIR}/adapter/driving/consumer/.gitkeep"

# Generate agent.md with minimal structure
generate_agent_md "$BC_NAME" "$BASE_DIR"

echo "✅ Bounded context '${BC_NAME}' created successfully at ${BASE_DIR}"
echo ""
echo "Structure created:"
echo "  - domain/                  (empty - ready for aggregates, entities, value objects)"
echo "  - application/command/     (ready for write operations)"
echo "  - application/query/       (ready for read operations)"
echo "  - adapter/driven/inmemory/ (empty - ready for repository implementations)"
echo "  - adapter/driving/http/    (empty - ready for HTTP handlers)"
echo "  - adapter/driving/consumer/ (ready for event consumers)"
echo "  - agent.md                 (minimal documentation structure generated)"
