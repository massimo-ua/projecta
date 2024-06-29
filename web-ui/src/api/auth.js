export class Auth {
  async getToken() {
    return localStorage.getItem('token');
  }
}
