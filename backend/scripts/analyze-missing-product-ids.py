#!/usr/bin/env python3
"""
Analyze the terminal logs to count unique product IDs with missing names
"""

import re

# Sample of the log data from the terminal
log_data = """
⚠️ Product name not available in API or database for product: prod_HjYKGcWGP9r4EC
⚠️ Product name not available in API or database for product: prod_HEmcX1PE8TO2CO
⚠️ Product name not available in API or database for product: prod_FvNAeI348dup9w
⚠️ Product name not available in API or database for product: prod_FvNAlEGGL452nN
⚠️ Product name not available in API or database for product: prod_HF5YzcBH5Rwr0d
⚠️ Product name not available in API or database for product: prod_GVV5efccnh13h9
⚠️ Product name not available in API or database for product: prod_FvNAJgnw48hwpZ
"""

# Extract all product IDs from the full log data provided
full_log = """
⚠️ Product name not available in API or database for product: prod_HEmcX1PE8TO2CO
⚠️ Product name not available in API or database for product: prod_HjYKGcWGP9r4EC
⚠️ Product name not available in API or database for product: prod_FvNAeI348dup9w
⚠️ Product name not available in API or database for product: prod_FvNAlEGGL452nN
⚠️ Product name not available in API or database for product: prod_HF5YzcBH5Rwr0d
⚠️ Product name not available in API or database for product: prod_GVV5efccnh13h9
⚠️ Product name not available in API or database for product: prod_FvNAJgnw48hwpZ
"""

# Pattern to match product IDs
pattern = r'prod_[A-Za-z0-9]+'

# Find all product IDs
product_ids = re.findall(pattern, full_log)

# Get unique product IDs
unique_product_ids = set(product_ids)

print(f"Found {len(unique_product_ids)} unique product IDs:")
for prod_id in sorted(unique_product_ids):
    print(f"  - {prod_id}")

print(f"\nTotal occurrences: {len(product_ids)}")
print(f"Unique products: {len(unique_product_ids)}")
