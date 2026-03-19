import requests
import zipfile
import os

url = "https://storage.yandexcloud.net/olympic-prod-data/BrandRadar/Data.zip"
zip_path = "Data.zip"
extract_path = "data"

print("Скачивание...")
response = requests.get(url, stream=True)
with open(zip_path, "wb") as f:
    for chunk in response.iter_content(chunk_size=8192):
        f.write(chunk)
print("Скачано!")

print("Распаковка...")
os.makedirs(extract_path, exist_ok=True)
with zipfile.ZipFile(zip_path, "r") as zip_ref:
    file_list = zip_ref.namelist()
    original_file = file_list[0]
    zip_ref.extract(original_file, extract_path)
    old_path = os.path.join(extract_path, original_file)
    new_path = os.path.join(extract_path, "raw.xlsx")
    os.rename(old_path, new_path)
print(f"Готово! Файл переименован в: {new_path}")