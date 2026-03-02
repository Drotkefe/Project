#!/usr/bin/env python3
"""
Test script for SexMoneyShare application
Tests the API endpoints and reports results
"""

import requests
import json
from datetime import datetime

BASE_URL = "http://localhost:8080"

def print_step(step_num, description):
    print(f"\n{'='*80}")
    print(f"STEP {step_num}: {description}")
    print('='*80)

def print_result(success, message):
    status = "✓ SUCCESS" if success else "✗ ERROR"
    print(f"{status}: {message}")

def api_call(method, endpoint, data=None):
    url = f"{BASE_URL}{endpoint}"
    headers = {'Content-Type': 'application/json'}
    
    try:
        if method == 'GET':
            response = requests.get(url, headers=headers)
        elif method == 'POST':
            response = requests.post(url, json=data, headers=headers)
        elif method == 'PUT':
            response = requests.put(url, json=data, headers=headers)
        elif method == 'DELETE':
            response = requests.delete(url, headers=headers)
        
        # Handle 204 No Content
        if response.status_code == 204:
            return True, None, "Success (No Content)"
        
        # Try to parse JSON response
        try:
            response_data = response.json()
        except:
            response_data = None
        
        if response.ok:
            return True, response_data, f"Success (Status {response.status_code})"
        else:
            error_msg = response_data.get('error', 'Request failed') if response_data else f"HTTP {response.status_code}"
            return False, response_data, error_msg
            
    except Exception as e:
        return False, None, str(e)

def main():
    print("\n" + "="*80)
    print("SEXMONEYSHARE APPLICATION TEST SUITE")
    print("="*80)
    
    # Store IDs for later use
    alice_id = None
    bob_id = None
    patrik_id = None
    carol_id = None
    dave_id = None
    trip_id = None
    payment_id = None
    
    # STEP 1: Navigate to Members tab (simulated - we'll just start testing members)
    print_step(1, "Click 'Members' tab")
    print_result(True, "Navigated to Members section (simulated)")
    
    # STEP 2: Add Alice
    print_step(2, "Click '+ Add Member' button, type 'Alice', click Save")
    success, data, msg = api_call('POST', '/members', {'name': 'Alice'})
    print_result(success, msg)
    if success and data:
        alice_id = data.get('id')
        print(f"   Alice ID: {alice_id}")
        print(f"   Response: {json.dumps(data, indent=2)}")
    
    # STEP 3: Add Bob
    print_step(3, "Click '+ Add Member' button, type 'Bob', click Save")
    success, data, msg = api_call('POST', '/members', {'name': 'Bob'})
    print_result(success, msg)
    if success and data:
        bob_id = data.get('id')
        print(f"   Bob ID: {bob_id}")
        print(f"   Response: {json.dumps(data, indent=2)}")
    
    # STEP 4: Edit Alice to Patrik
    print_step(4, "Click 'Edit' on Alice's card, change name to 'Patrik', click Save")
    if alice_id:
        success, data, msg = api_call('PUT', f'/members/{alice_id}', {'name': 'Patrik'})
        print_result(success, msg)
        if success:
            patrik_id = alice_id  # Same ID, just renamed
            print(f"   Patrik ID: {patrik_id}")
            if data:
                print(f"   Response: {json.dumps(data, indent=2)}")
    else:
        print_result(False, "Alice ID not available from previous step")
    
    # STEP 5: Delete Patrik
    print_step(5, "Click 'Delete' on Patrik's card, confirm the dialog")
    if patrik_id:
        success, data, msg = api_call('DELETE', f'/members/{patrik_id}')
        print_result(success, msg)
        if data:
            print(f"   Response: {json.dumps(data, indent=2)}")
    else:
        print_result(False, "Patrik ID not available from previous step")
    
    # STEP 6: Add Alice, Carol, and Dave
    print_step(6, "Click '+ Add Member' again, add 'Alice', 'Carol', and 'Dave'")
    
    # Add Alice back
    print("   Adding Alice...")
    success, data, msg = api_call('POST', '/members', {'name': 'Alice'})
    print_result(success, f"Alice: {msg}")
    if success and data:
        alice_id = data.get('id')
        print(f"      Alice ID: {alice_id}")
    
    # Add Carol
    print("   Adding Carol...")
    success, data, msg = api_call('POST', '/members', {'name': 'Carol'})
    print_result(success, f"Carol: {msg}")
    if success and data:
        carol_id = data.get('id')
        print(f"      Carol ID: {carol_id}")
    
    # Add Dave
    print("   Adding Dave...")
    success, data, msg = api_call('POST', '/members', {'name': 'Dave'})
    print_result(success, f"Dave: {msg}")
    if success and data:
        dave_id = data.get('id')
        print(f"      Dave ID: {dave_id}")
    
    # Get all members to confirm
    print("\n   Current members:")
    success, data, msg = api_call('GET', '/members')
    if success and data:
        for member in data:
            print(f"      - {member['name']} (ID: {member['id']})")
    
    # STEP 7: Navigate to Trips tab
    print_step(7, "Click 'Trips' tab")
    print_result(True, "Navigated to Trips section (simulated)")
    
    # STEP 8: Create Beach Trip
    print_step(8, "Click '+ New Trip', fill in details, click Save")
    today = datetime.now().strftime('%Y-%m-%d')
    member_ids = [mid for mid in [alice_id, bob_id, carol_id, dave_id] if mid is not None]
    
    trip_data = {
        'name': 'Beach Trip',
        'total_cost': 1000,
        'date': today,
        'member_ids': member_ids
    }
    print(f"   Trip data: name='Beach Trip', cost=1000, date={today}")
    print(f"   Members: {member_ids}")
    
    success, data, msg = api_call('POST', '/trips', trip_data)
    print_result(success, msg)
    if success and data:
        trip_id = data.get('id')
        print(f"   Trip ID: {trip_id}")
        print(f"   Response: {json.dumps(data, indent=2)}")
    
    # STEP 9: Add payment for Alice
    print_step(9, "Expand trip, click '+ Add Payment', select Alice, amount=1000, click Save")
    if trip_id and alice_id:
        payment_data = {
            'member_id': alice_id,
            'amount': 1000
        }
        success, data, msg = api_call('POST', f'/trips/{trip_id}/payments', payment_data)
        print_result(success, msg)
        if success and data:
            payment_id = data.get('id')
            print(f"   Payment ID: {payment_id}")
            print(f"   Response: {json.dumps(data, indent=2)}")
    else:
        print_result(False, "Trip ID or Alice ID not available")
    
    # STEP 10: Edit trip cost to 1200
    print_step(10, "Click 'Edit Trip', change cost to 1200, click Save")
    if trip_id:
        updated_trip_data = {
            'name': 'Beach Trip',
            'total_cost': 1200,
            'date': today,
            'member_ids': member_ids
        }
        success, data, msg = api_call('PUT', f'/trips/{trip_id}', updated_trip_data)
        print_result(success, msg)
        if data:
            print(f"   Response: {json.dumps(data, indent=2)}")
    else:
        print_result(False, "Trip ID not available")
    
    # STEP 11: Delete trip
    print_step(11, "Click 'Delete Trip'")
    if trip_id:
        success, data, msg = api_call('DELETE', f'/trips/{trip_id}')
        print_result(success, msg)
        if data:
            print(f"   Response: {json.dumps(data, indent=2)}")
    else:
        print_result(False, "Trip ID not available")
    
    # STEP 12: Check Dashboard
    print_step(12, "Click 'Dashboard' tab and check balances and settlements")
    print_result(True, "Navigated to Dashboard section (simulated)")
    
    # Get balances
    print("\n   Fetching balances...")
    success, data, msg = api_call('GET', '/balances')
    if success and data:
        print_result(True, "Balances retrieved")
        
        balances = data.get('balances', [])
        settlements = data.get('settlements', [])
        
        print("\n   BALANCES:")
        if balances:
            for bal in balances:
                member_name = bal.get('member', {}).get('name', 'Unknown')
                balance = bal.get('balance', 0)
                sign = '+' if balance > 0 else ''
                status = 'positive' if balance > 0.01 else 'negative' if balance < -0.01 else 'zero'
                print(f"      {member_name}: {sign}{balance:.2f} HUF ({status})")
        else:
            print("      No balances found (or no members)")
        
        print("\n   SETTLEMENTS:")
        if settlements:
            for settlement in settlements:
                from_name = settlement.get('from_name', 'Unknown')
                to_name = settlement.get('to_name', 'Unknown')
                amount = settlement.get('amount', 0)
                print(f"      {from_name} → {to_name}: {amount:.2f} HUF")
        else:
            print("      All settled up! (or no settlements needed)")
    else:
        print_result(False, f"Failed to retrieve balances: {msg}")
    
    print("\n" + "="*80)
    print("TEST SUITE COMPLETED")
    print("="*80 + "\n")

if __name__ == '__main__':
    main()
