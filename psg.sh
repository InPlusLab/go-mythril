# rm log/*
rm go-mythril
go build -o go-mythril main3.go

i=0
rlimit=2000000
maxRLimit=2000000
codeList=("","","")
for line in `cat top10-six1.txt |tr -d ' '| tr -d '\r'`
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
    ./go-mythril -goFuncCount 1  -rLimit $rlimit -maxRLimit $maxRLimit -contractName "${codeList[0]}" -creationCode "${codeList[1]}" -runtimeCode "${codeList[2]}" -skipTimes 0 -index 0 > ./lognew/$rlimit-$maxRLimit-1-"${codeList[0]}".log &

    i=0
  fi
done
