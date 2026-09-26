export class InvitationsRepository {
  #request;

  constructor(request) {
    this.#request = request;
  }

  async create(email) {
    return this.#request.post('/invitations', { email });
  }

  async list() {
    const response = await this.#request.get('/invitations');
    return response?.invitations || [];
  }

  async delete(id) {
    return this.#request.delete(`/invitations/${id}`);
  }

  async validate(code) {
    return this.#request.get(`/invitations/code/${code}`);
  }
}

export default InvitationsRepository;
