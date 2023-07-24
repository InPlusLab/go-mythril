rm log/*
rm go-mythril
go build -o go-mythril main3.go

i=0
codeList=("","","")
for line in `cat top100-one.txt |tr -d ' '| tr -d '\r'`
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
    ./go-mythril -goFuncCount 16 -maxRLimit 10000000 -rLimit 2000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-16a.log
    # ./go-mythril -goFuncCount 8 -maxRLimit 10000000 -rLimit 2000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-8a.log
    # ./go-mythril -goFuncCount 4 -maxRLimit 10000000 -rLimit 2000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-4a.log
    # ./go-mythril -goFuncCount 2 -maxRLimit 10000000 -rLimit 2000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-2a.log
    # ./go-mythril -goFuncCount 1 -maxRLimit 10000000 -rLimit 2000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-1a.log
    
    # ./go-mythril -goFuncCount 16 -maxRLimit 10000000 -rLimit 2000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-16b.log
    # ./go-mythril -goFuncCount 8 -maxRLimit 10000000 -rLimit 2000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-8b.log
    # ./go-mythril -goFuncCount 4 -maxRLimit 10000000 -rLimit 2000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-4b.log
    # ./go-mythril -goFuncCount 2 -maxRLimit 10000000 -rLimit 2000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-2b.log
    # ./go-mythril -goFuncCount 1 -maxRLimit 10000000 -rLimit 2000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-1b.log


    # ./go-mythril -goFuncCount 16 -maxRLimit 5000000 -rLimit 1000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-16a.log
    # ./go-mythril -goFuncCount 16 -maxRLimit 5000000 -rLimit 1000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-16b.log
    # ./go-mythril -goFuncCount 16 -maxRLimit 5000000 -rLimit 1000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-16c.log
    # ./go-mythril -goFuncCount 16 -maxRLimit 5000000 -rLimit 1000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-16d.log
    # ./go-mythril -goFuncCount 16 -maxRLimit 5000000 -rLimit 1000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-16e.log
    # ./go-mythril -goFuncCount 16 -maxRLimit 5000000 -rLimit 1000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-16f.log
    # ./go-mythril -goFuncCount 16 -maxRLimit 5000000 -rLimit 1000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-16g.log
    # ./go-mythril -goFuncCount 16 -maxRLimit 5000000 -rLimit 1000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-16h.log

    # ./go-mythril -goFuncCount 1 -maxRLimit 5000000 -rLimit 1000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-1a.log &
    # ./go-mythril -goFuncCount 1 -maxRLimit 5000000 -rLimit 1000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-1b.log &
    # ./go-mythril -goFuncCount 1 -maxRLimit 5000000 -rLimit 1000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-1c.log &
    # ./go-mythril -goFuncCount 1 -maxRLimit 5000000 -rLimit 1000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-1d.log &
    # ./go-mythril -goFuncCount 1 -maxRLimit 5000000 -rLimit 1000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-1e.log &
    # ./go-mythril -goFuncCount 1 -maxRLimit 5000000 -rLimit 1000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-1f.log &
    # ./go-mythril -goFuncCount 1 -maxRLimit 5000000 -rLimit 1000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-1g.log &
    # ./go-mythril -goFuncCount 1 -maxRLimit 5000000 -rLimit 1000000 -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./log/"${codeList[0]}"-1h.log

    i=0
  fi
done
