FROM python:3.8

COPY . /code
WORKDIR /code

RUN pip install -r requirements.txt

EXPOSE 9091
CMD ["python", "-u",  "field.py"]
