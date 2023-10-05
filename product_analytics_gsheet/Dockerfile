FROM python:3.9-alpine

RUN mkdir -p /usr/local/var/product

WORKDIR /usr/local/var/product

COPY . ./

ENV PYTHONPATH /usr/local/var/product
ENV TZ=Asia/Kolkata

# Install Python dependencies
RUN pip install -r ./packages_requirements.txt

RUN chmod +x ./write_to_sheet.py

CMD ["python3", "/usr/local/var/product/write_to_sheet.py"]