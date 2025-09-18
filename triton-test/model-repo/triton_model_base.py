import logging
import os
import triton_python_backend_utils as pb_utils

class TritonPythonModelBase:
    """
    Base class for Triton Python models with logging support.
    Provides standardized logging configuration across all models.
    """
    
    def __init__(self):
        self.logger = None
    
    def _setup_logging(self, model_name):
        """
        Setup logging configuration for the model.
        
        Args:
            model_name (str): Name of the model for logger identification
        """
        # Get log level from environment variable, default to INFO
        log_level = os.getenv('TRITON_LOG_LEVEL', 'INFO').upper()
        
        # Create logger with model-specific name
        self.logger = logging.getLogger(f'triton.{model_name}')
        
        # Avoid duplicate handlers if logger already exists
        if not self.logger.handlers:
            # Create console handler
            handler = logging.StreamHandler()
            
            # Create formatter matching Triton's native log format
            # Format: I0910 09:15:41.294772 1 filename:line] "message"
            formatter = logging.Formatter(
                '%(levelname).1s%(asctime)s.%(msecs)03d000 - %(name)s] "%(message)s"',
                datefmt='%m%d %H:%M:%S'
            )
            handler.setFormatter(formatter)
            
            # Add handler to logger
            self.logger.addHandler(handler)
            
            # Set log level
            try:
                level = getattr(logging, log_level)
                self.logger.setLevel(level)
                handler.setLevel(level)
            except AttributeError:
                # Fallback to INFO if invalid level provided
                self.logger.setLevel(logging.INFO)
                handler.setLevel(logging.INFO)
                self.logger.warning(f"Invalid log level '{log_level}', using INFO instead")
        
        # Prevent propagation to avoid duplicate logs
        self.logger.propagate = False
        
        self.logger.info(f"{model_name} model logger initialized with level {self.logger.level}")