export class BaseStateClass {
  getCopy = () => ({ ...this });

  setItem = (key: keyof this, val: any) => {
    this[key] = val;
  };
}
