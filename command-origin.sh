sudo go build -o go-mythril main3.go

i=0
index=0
codeList=("","","")
for line in `cat contractCode-left.txt |tr -d ' '| tr -d '\r'`
do
  codeList[i]=$line
  # shellcheck disable=SC2006
  # shellcheck disable=SC2003
  i=`expr $i + 1`
  if [ $i == 3 ]
  then
    i=0
    # shellcheck disable=SC2006
    index=`expr $index + 1`
  fi
  if [ $index == $1 ]
  then
    ./go-mythril -goFuncCount 1 -maxRLimit 1008610086 -rLimit 5000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" > /home/zhangn279/log/smartbugs/500w-1goFuncCount-noSkip/"${codeList[0]}".log
    break
  fi
done