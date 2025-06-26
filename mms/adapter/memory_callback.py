from transformers import TrainerCallback
import torch
import psutil

class MemoryCallback(TrainerCallback):
    def on_step_end(self, args, state, control, **kwargs):
        if state.global_step % 10 == 0:  # Log every 10 steps
            # GPU memory
            if torch.cuda.is_available():
                gpu_mem = torch.cuda.memory_allocated() / 1024**3
                gpu_max = torch.cuda.max_memory_allocated() / 1024**3
                print(f"Step {state.global_step}: GPU {gpu_mem:.2f}GB (max: {gpu_max:.2f}GB)")
            # CPU memory
            cpu_mem = psutil.virtual_memory().percent
            print(f"Step {state.global_step}: CPU memory {cpu_mem:.1f}%")