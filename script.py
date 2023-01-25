import os
import multiprocessing

cpu_count = int(multiprocessing.cpu_count()/3)
benchmark_process_lock = multiprocessing.Semaphore(cpu_count)

def run_smartbugs(index):
    ret = os.fork()
    if ret != 0:
        return ret
    else:
        benchmark_process_lock.acquire()

    cmd = f"./command-origin.sh {index}"
    os.system(cmd)

    benchmark_process_lock.release()
    exit()

pid_cnt = 0
for index in range(1,3):
    run_smartbugs(index)
    pid_cnt += 1

while pid_cnt >0:
    pid, status = os.wait()
    pid_cnt -= 1