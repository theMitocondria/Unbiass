import json
import threading
import time
from openai import OpenAI
from prompt import system
import sys
import os
from dotenv import load_dotenv
import redis 

r = redis.Redis(host='localhost', port=6379, db=0)
load_dotenv()
api_key = os.getenv("OPENAI_API_KEY")

if len(sys.argv) < 4:
    print("Usage: python script.py <input_file.json>")
    sys.exit(1)

input_file = sys.argv[1]
transaction_id = sys.argv[3]

with open(input_file, "r", encoding="utf-8") as f:
    code_data = json.load(f)

openai = OpenAI(
    api_key= api_key,
    base_url="https://api.deepinfra.com/v1/openai",
)

def create_chat_instance(api_key):
    return openai.chat.completions.create

chat_completion = create_chat_instance(api_key)

system_prompt = {"role": "system", "content": system}
human_prompt_template = "Code 1: \n\n{code1}\n \n\nCode 2: \n\n{code2}\n"
cheaters = set()
non_cheaters = set()
totalDesiredCandidates =  int(sys.argv[2])
students_done = 0
lock = threading.Lock()

def execute_chain(i, j, code1, code2, cheaters):
    human_prompt = {"role": "user", "content": human_prompt_template.format(code1=code1, code2=code2)}
    response = chat_completion(
        model="Qwen/Qwen2.5-Coder-32B-Instruct",
        temperature=0.5,
        messages=[system_prompt, human_prompt]
    )
    # Extract the content from the response
    response_content = response.choices[0].message.content.strip()

    if response_content.startswith("{{") and response_content.endswith("}}"):
        response_content = response_content[1:-1]  
   
    try:
        score = json.loads(response_content).get("score", 0)
    except json.JSONDecodeError:
        print(f"Failed to decode response: {response_content}")
        return

    if score > 0.7:
        with lock :
            cheaters.add(i)
            cheaters.add(j)
            global students_done
            students_done+= 1
            progress = students_done/totalDesiredCandidates * 100
            r.set(f"progress:{transaction_id}", int(progress))
            # print(progress)



for i in range(0, len(code_data['submissions'])):  # Note: This loop only iterates once as per your example. Adjust the range for more iterations.
    
    if len(non_cheaters) >= totalDesiredCandidates:
        break

    if i in cheaters:
        continue
    
    for questions in range (0 , code_data['no_of_questions'] ):
        threads = []
        for j in range(i + 1, len(code_data['submissions'])):
            if j not in cheaters:
                thread = threading.Thread(target=execute_chain, args=(i, j, code_data['submissions'][i][questions]['Code'], code_data['submissions'][j][questions]['Code'], cheaters))
                threads.append(thread)
                if len(threads) >= 200 :
                    for thread in threads:
                        thread.start()
                    for thread in threads:
                        thread.join()
                    threads = []    

        for thread in threads:
            thread.start()
        for thread in threads:
            thread.join()

        if i not in cheaters:
            non_cheaters.add(code_data['submissions'][i][questions]['ID']) 
            with lock :
                students_done+=1
                progress = students_done/totalDesiredCandidates * 100
                r.set(f"progress:{transaction_id}", int(progress))
    
r.set(f"progress:{transaction_id}", int("100"))
print(cheaters)
