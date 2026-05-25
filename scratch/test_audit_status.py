import urllib.request
import urllib.parse
import json

def main():
    # 1. Login to get token
    login_url = "http://127.0.0.1:8080/api/v1/auth/login"
    login_data = json.dumps({
        "email": "testadmin@example.com",
        "password": "password"
    }).encode('utf-8')

    req = urllib.request.Request(
        login_url,
        data=login_data,
        headers={'Content-Type': 'application/json'}
    )

    try:
        with urllib.request.urlopen(req) as response:
            resp_data = json.loads(response.read().decode('utf-8'))
            token = resp_data['data']['token']
            print(f"Logged in successfully. Token length: {len(token)}")
    except Exception as e:
        print(f"Login failed: {e}")
        return

    # Use the hotel fee declaration ID from query_declarations
    decl_id = "065f18d2-5500-415c-bf3c-5a67f72c5a92"
    audit_url = f"http://127.0.0.1:8080/api/v1/admin/tax-new/declare/{decl_id}/audit-status"
    
    audit_data = json.dumps({
        "status": "audit_failed",
        "notes": "จำนวนเตียงและรายรับของโรงแรมไม่สอดคล้องกับรายงานสรุปในเอกสารแนบ"
    }).encode('utf-8')

    req_audit = urllib.request.Request(
        audit_url,
        data=audit_data,
        headers={
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {token}'
        }
    )

    try:
        with urllib.request.urlopen(req_audit) as response:
            print("Audit Status Update Succeeded!")
            print(response.read().decode('utf-8'))
    except Exception as e:
        print(f"Audit Status Update Failed: {e}")
        if hasattr(e, 'read'):
            print(e.read().decode('utf-8'))

if __name__ == '__main__':
    main()
