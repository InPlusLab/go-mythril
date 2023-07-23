sudo go build -o go-mythril main3.go

i=0
codeList=("","","")
for line in `cat top100-contractCode-left.txt |tr -d ' '| tr -d '\r'`
do
  codeList[i]=$line
  # shellcheck disable=SC2006
  # shellcheck disable=SC2003
  i=`expr $i + 1`
  if [ $i == 3 ]
  then
    # shellcheck disable=SC2034
    # shellcheck disable=SC2006
    # shellcheck disable=SC1116
    # shellcheck disable=SC2154
    ./go-mythril -goFuncCount 16 -maxRLimit 9600000 -rLimit 3200000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 5 -index 0 > ~/log/top100/"${codeList[0]}".log
    i=0
  fi
done
