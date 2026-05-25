import urllib.request
import urllib.parse
import json
import os

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
        if hasattr(e, 'read'):
            print(e.read().decode('utf-8'))
        return

    # Helper function to perform multipart upload using urllib
    def upload_file(url, file_path):
        boundary = b'----WebKitFormBoundary7MA4YWxkTrZu0gW'
        filename = os.path.basename(file_path)
        
        with open(file_path, 'rb') as f:
            file_content = f.read()
            
        body = (
            b'--' + boundary + b'\r\n' +
            b'Content-Disposition: form-data; name="file"; filename="' + filename.encode('utf-8') + b'"\r\n' +
            b'Content-Type: application/octet-stream\r\n\r\n' +
            file_content + b'\r\n' +
            b'--' + boundary + b'--\r\n'
        )
        
        req = urllib.request.Request(
            url,
            data=body,
            headers={
                'Content-Type': f'multipart/form-data; boundary={boundary.decode("utf-8")}',
                'Authorization': f'Bearer {token}'
            }
        )
        
        try:
            with urllib.request.urlopen(req) as response:
                print(f"Upload to {url} Succeeded!")
                print(response.read().decode('utf-8'))
        except Exception as e:
            print(f"Upload to {url} Failed: {e}")
            if hasattr(e, 'read'):
                print(e.read().decode('utf-8'))

    # 2. Upload KTB Reconciliation file
    ktb_path = r"C:\Users\phnjk\.gemini\antigravity-ide\brain\6d9cd73e-a39d-4db6-9b99-1e51cf80f4bf\scratch\ktb_reconciliation_sample.txt"
    print("Uploading KTB reconciliation file...")
    upload_file("http://127.0.0.1:8080/api/v1/admin/tax-new/reconciliation/upload", ktb_path)

    # 3. Upload e-LAAS Daily Summary file
    elaas_path = r"C:\Users\phnjk\.gemini\antigravity-ide\brain\6d9cd73e-a39d-4db6-9b99-1e51cf80f4bf\scratch\elaas_daily_summary_sample.csv"
    print("Uploading e-LAAS daily summary file...")
    upload_file("http://127.0.0.1:8080/api/v1/admin/tax-new/elaas/upload", elaas_path)

if __name__ == '__main__':
    main()
