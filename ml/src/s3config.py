import os
import dotenv

class S3Config:
    S3_ENDPOINT_URL: str = None
    S3_ACCESS_KEY: str = None
    S3_SECRET_KEY: str = None
    S3_REGION_NAME: str = None

    def __init__(self):
        dotenv.load_dotenv()

        self.S3_ENDPOINT_URL = os.getenv("S3_ENDPOINT_URL", self.S3_ENDPOINT_URL)
        self.S3_ACCESS_KEY = os.getenv("S3_ACCESS_KEY", self.S3_ACCESS_KEY)
        self.S3_SECRET_KEY = os.getenv("S3_SECRET_KEY", self.S3_SECRET_KEY)
        self.S3_REGION_NAME = os.getenv("REGION_NAME", self.S3_REGION_NAME)