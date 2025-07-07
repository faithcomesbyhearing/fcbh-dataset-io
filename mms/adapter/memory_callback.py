from transformers import TrainerCallback
import torch
import psutil

class MemoryCallback(TrainerCallback):
    def on_step_end(self, args, state, control, **kwargs):
        if torch.cuda.is_available():
            allocated = torch.cuda.memory_allocated() / 1024**3
            if allocated > 30:  # If over 30GB, force cleanup
                print(f"Empty Cache: Step {state.global_step}: GPU {gpu_mem:.2f}GB (max: {gpu_max:.2f}GB)")
                torch.cuda.empty_cache()
        if state.global_step % 1 == 0:  # Log every 1 steps
            if torch.cuda.is_available():
                gpu_mem = torch.cuda.memory_allocated() / 1024**3
                gpu_max = torch.cuda.max_memory_allocated() / 1024**3
                print(f"Step {state.global_step}: GPU {gpu_mem:.2f}GB (max: {gpu_max:.2f}GB)")
            cpu_mem = psutil.virtual_memory().percent
            print(f"Step {state.global_step}: CPU memory {cpu_mem:.1f}%")