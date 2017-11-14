aws cloudformation describe-stacks --stack-name efs-testing-readwrite
if [ $? == 0 ]; then 
  action=update-stack; 
else 
  action=create-stack; 
fi; 
vpc="`aws ec2 describe-vpcs --filters Name=isDefault,Values=true --query Vpcs[0].VpcId --output text`"
subnets="`aws ec2 describe-subnets --filters Name=vpc-id,Values=$vpc --query Subnets[*][SubnetId] --output text | tr '\n' ',' | sed 's/,$//'`"
cat > params.json <<-eof
{ "VPC": "$vpc", "Subnets": "$subnets" }
eof
cat params.json
aws cloudformation $action --template-body file:///code/stack.json --stack-name efs-testing-readwrite --region us-east-1 --capabilities CAPABILITY_IAM --parameters "ParameterKey=VPC,ParameterValue=$vpc" "ParameterKey=Subnets,ParameterValue=\"$subnets\""
