sudo go build -o go-mythril main3.go

i=0
codeList=("","","")
for line in `cat contractCode.txt |tr -d ' '| tr -d '\r'`
do
  codeList[i]=$line
  # shellcheck disable=SC2006
  # shellcheck disable=SC2003
  i=`expr $i + 1`
  if [ $i == 3 ]
  then
    ./go-mythril -goFuncCount 8 -maxRLimit 1008610086 -rLimit 5000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" > /home/codepatient/log/smartbugs/500w-8goFuncCount-noSkip/"${codeList[0]}".log
    i=0
  fi
done



