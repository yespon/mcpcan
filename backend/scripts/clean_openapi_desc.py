#!/usr/bin/env python3
"""
Clean @inject_tag annotations from OpenAPI documentation, extracting only desc: content.
Usage: python3 scripts/clean_openapi_desc.py <openapi.yaml> [--title TITLE] [--version VERSION] [--server URL]
"""

import re
import sys
import yaml

def clean_inject_tag(text):
    """Extract desc: content from @inject_tag annotation"""
    if not text or '@inject_tag' not in text:
        return text
    
    # Match desc:"content" pattern
    match = re.search(r'desc:"([^"]+)"', text)
    if match:
        return match.group(1)
    
    # If no desc:, remove entire @inject_tag
    cleaned = re.sub(r'@inject_tag:[^"]*("[^"]*"\s*)*', '', text).strip()
    return cleaned if cleaned else text

def process_schema(schema):
    """Recursively process description in schema"""
    if isinstance(schema, dict):
        if 'description' in schema:
            schema['description'] = clean_inject_tag(schema['description'])
        for key, value in schema.items():
            if isinstance(value, (dict, list)):
                process_schema(value)
    elif isinstance(schema, list):
        for item in schema:
            process_schema(item)

def process_openapi(file_path, title=None, version=None, server_url=None):
    """Process OpenAPI file"""
    with open(file_path, 'r', encoding='utf-8') as f:
        doc = yaml.safe_load(f)
    
    # Update info metadata
    if 'info' not in doc:
        doc['info'] = {}
    
    if title:
        doc['info']['title'] = title
    if version:
        doc['info']['version'] = version
    
    # Add/update servers
    if server_url:
        doc['servers'] = [{'url': server_url, 'description': 'API Server'}]
    
    # Process paths
    if 'paths' in doc:
        for path, methods in doc['paths'].items():
            for method, details in methods.items():
                if isinstance(details, dict):
                    # Process parameter descriptions
                    if 'parameters' in details:
                        for param in details['parameters']:
                            if 'description' in param:
                                param['description'] = clean_inject_tag(param['description'])
                    # Process request body
                    process_schema(details.get('requestBody', {}))
                    # Process responses
                    process_schema(details.get('responses', {}))
    
    # Process components/schemas
    if 'components' in doc and 'schemas' in doc['components']:
        for schema_name, schema in doc['components']['schemas'].items():
            process_schema(schema)
    
    # Write back to file
    with open(file_path, 'w', encoding='utf-8') as f:
        yaml.dump(doc, f, allow_unicode=True, default_flow_style=False, sort_keys=False)
    
    print(f"✅ Cleaned: {file_path}")

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Usage: python3 clean_openapi_desc.py <openapi.yaml> [--title TITLE] [--version VERSION] [--server URL]")
        sys.exit(1)
    
    file_path = sys.argv[1]
    
    # Parse optional arguments
    title = None
    version = None
    server_url = None
    
    i = 2
    while i < len(sys.argv):
        if sys.argv[i] == '--title' and i + 1 < len(sys.argv):
            title = sys.argv[i + 1]
            i += 2
        elif sys.argv[i] == '--version' and i + 1 < len(sys.argv):
            version = sys.argv[i + 1]
            i += 2
        elif sys.argv[i] == '--server' and i + 1 < len(sys.argv):
            server_url = sys.argv[i + 1]
            i += 2
        else:
            i += 1
    
    process_openapi(file_path, title=title, version=version, server_url=server_url)
