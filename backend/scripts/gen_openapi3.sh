#!/bin/bash
# Generate OpenAPI 3 documentation and output to mcpcan-openapi.json
# Usage: ./scripts/gen_openapi3.sh

set -e

cd "$(dirname "$0")/.."

# Configuration
TITLE="MCPCAN OpenAPI"
SERVER_URL="https://DOMAIN/api/market"
VERSION_FILE="../VERSION"
OUTPUT_YAML="api/openapi.yaml"
OUTPUT_JSON="init-data/openapi-file/mcpcan-openapi.json"

# Read version
if [ -f "$VERSION_FILE" ]; then
    VERSION=$(cat "$VERSION_FILE" | tr -d '\n')
    echo "📦 Version: $VERSION"
else
    VERSION="v1.0.0"
    echo "⚠️  VERSION file not found, using default: $VERSION"
fi

echo "📄 Generating OpenAPI 3 documentation..."
echo "   Title: $TITLE"
echo "   Server: $SERVER_URL"

# 1. Generate with buf
cd api
buf generate
cd ..

# 2. Check generated file
if [ ! -f "$OUTPUT_YAML" ]; then
    echo "❌ File not found: $OUTPUT_YAML"
    exit 1
fi

echo "📁 Found OpenAPI file: $OUTPUT_YAML"

# 3. Clean @inject_tag and set metadata
echo "🧹 Cleaning @inject_tag descriptions and setting metadata..."
if command -v python3 &> /dev/null; then
    python3 scripts/clean_openapi_desc.py "$OUTPUT_YAML" \
        --title "$TITLE" \
        --version "$VERSION" \
        --server "$SERVER_URL"
else
    echo "⚠️  python3 not installed, skipping description cleanup"
fi

# 4. Convert to JSON and output to target location
echo "📝 Converting to JSON format..."
if command -v python3 &> /dev/null; then
    python3 -c "
import yaml
import json
with open('$OUTPUT_YAML', 'r', encoding='utf-8') as f:
    doc = yaml.safe_load(f)
with open('$OUTPUT_JSON', 'w', encoding='utf-8') as f:
    json.dump(doc, f, ensure_ascii=False, indent=2)
print('✅ Output: $OUTPUT_JSON')
"
else
    echo "⚠️  python3 not installed, cannot convert to JSON"
fi

# 5. Show statistics
if [ -f "$OUTPUT_JSON" ]; then
    PATHS_COUNT=$(grep -c '"/' "$OUTPUT_JSON" 2>/dev/null || echo "0")
    echo ""
    echo "📊 Documentation statistics:"
    echo "   YAML: $OUTPUT_YAML"
    echo "   JSON: $OUTPUT_JSON"
    echo "   Size: $(ls -lh "$OUTPUT_JSON" | awk '{print $5}')"
    echo "   Paths: ~$((PATHS_COUNT / 2))"
fi

echo ""
echo "✅ OpenAPI 3 documentation generation complete!"
