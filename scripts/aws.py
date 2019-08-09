from troposphere import ec2, Output
from troposphere import Template, Ref, Tags

class Lambda():
    def __init__(self):
        self.template = Template()

    @staticmethod
    def template():
        return Lambda().template.to_json()


if __name__ == '__main__':
    print(Lambda.template())