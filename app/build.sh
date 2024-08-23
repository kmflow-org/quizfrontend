if [ $# -eq 0 ]; then
    echo "Error: No argument provided."
    echo "Usage: $0 <version>"
    exit 1
fi

GOOS=linux GOARCH=amd64 go build
rm -rf release || 1
mkdir release
cp -r quizzes release
cp -r templates release
cp config.yaml release
mv quizengine release
tar -czvf release-$1.tar.gz release
aws s3 cp release-$1.tar.gz s3://kmflow-org-artifacts/  --profile kmflowdev --region us-west-2
