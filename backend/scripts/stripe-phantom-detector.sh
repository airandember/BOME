#!/bin/bash
# 🔍 STRIPE PHANTOM DETECTOR
# This script will query Stripe API directly to confirm ghost products

echo "🔍 CHECKING STRIPE API FOR GHOST PRODUCTS..."
echo "================================================"

# Your known ghost products
GHOST_PRODUCTS=(
    "prod_HjYKGcWGP9r4EC"
    "prod_HEmcX1PE8TO2CO" 
    "prod_FvNAeI348dup9w"
    "prod_FvNAlEGGL452nN"
    "prod_HF5YzcBH5Rwr0d"
    "prod_GVV5efccnh13h9"
    "prod_FvNAJgnw48hwpZ"
)

# Replace with your actual Stripe secret key
STRIPE_SK="sk_live_YOUR_KEY_HERE"

for product_id in "${GHOST_PRODUCTS[@]}"; do
    echo "🔍 Checking $product_id..."
    
    response=$(curl -s -u "$STRIPE_SK:" \
        "https://api.stripe.com/v1/products/$product_id")
    
    if echo "$response" | grep -q "No such product"; then
        echo "❌ $product_id: DOES NOT EXIST (404)"
    elif echo "$response" | grep -q '"active": false'; then
        echo "👻 $product_id: EXISTS but INACTIVE (phantom)"
    elif echo "$response" | grep -q '"active": true'; then
        echo "✅ $product_id: EXISTS and ACTIVE"
    else
        echo "❓ $product_id: UNKNOWN STATUS"
    fi
    echo "---"
done

echo "🎯 CONCLUSION:"
echo "If products show '404 DOES NOT EXIST' = True ghosts from old account"
echo "If products show 'INACTIVE' = Soft-deleted products still in API"
echo "If products show 'ACTIVE' = Something is wrong with your dashboard view"
