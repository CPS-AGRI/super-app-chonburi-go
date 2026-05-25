#!/bin/sh
set -eu

: "${MINIO_ROOT_USER:?MINIO_ROOT_USER is required}"
: "${MINIO_ROOT_PASSWORD:?MINIO_ROOT_PASSWORD is required}"
: "${MINIO_BUCKET:?MINIO_BUCKET is required}"
: "${MINIO_APP_USER:?MINIO_APP_USER is required}"
: "${MINIO_APP_PASSWORD:?MINIO_APP_PASSWORD is required}"
: "${MINIO_REGION:?MINIO_REGION is required}"
: "${MINIO_POLICY_NAME:?MINIO_POLICY_NAME is required}"
MINIO_PUBLIC_READ="${MINIO_PUBLIC_READ:-false}"

endpoint="http://minio:9000"

mc alias set local "$endpoint" "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"

echo "Waiting for MinIO to be ready..."
until mc ready local >/dev/null 2>&1; do
  sleep 2
done

echo "Creating bucket if it does not exist..."
mc mb --ignore-existing --region "$MINIO_REGION" "local/$MINIO_BUCKET"

policy_file="/tmp/${MINIO_POLICY_NAME}.json"
cat > "$policy_file" <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads"
      ],
      "Resource": [
        "arn:aws:s3:::$MINIO_BUCKET"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject",
        "s3:AbortMultipartUpload",
        "s3:ListMultipartUploadParts"
      ],
      "Resource": [
        "arn:aws:s3:::$MINIO_BUCKET/*"
      ]
    }
  ]
}
EOF

echo "Creating or updating bucket-scoped policy..."
mc admin policy create local "$MINIO_POLICY_NAME" "$policy_file" >/dev/null

if ! mc admin user info local "$MINIO_APP_USER" >/dev/null 2>&1; then
  echo "Creating app user..."
  mc admin user add local "$MINIO_APP_USER" "$MINIO_APP_PASSWORD"
else
  echo "App user already exists, skipping creation..."
fi

echo "Attaching policy to app user..."
mc admin policy attach local "$MINIO_POLICY_NAME" --user "$MINIO_APP_USER" >/dev/null

if [ "$MINIO_PUBLIC_READ" = "true" ] || [ "$MINIO_PUBLIC_READ" = "1" ]; then
  echo "Enabling public read access for bucket objects..."
  mc anonymous set download "local/$MINIO_BUCKET"
else
  echo "Keeping bucket private..."
  mc anonymous set none "local/$MINIO_BUCKET"
fi

echo "MinIO bootstrap complete."
