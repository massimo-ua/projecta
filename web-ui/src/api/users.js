import { User } from '../models/User.js';

const toDomain = (userDto) => {
  if (!userDto) return null;
  return new User({
    id: userDto.customer_id,
    customerId: userDto.customer_id,
    firstName: userDto.first_name,
    lastName: userDto.last_name,
    displayName: userDto.display_name,
    roles: userDto.roles || [],
  });
};

const toAssignRolesDTO = (roles) => ({
  roles: Array.isArray(roles) ? roles : [],
});

export class UsersRepository {
  #request;

  constructor(request) {
    this.#request = request;
  }

  async getUsers(limit = 10, offset = 0) {
    const query = new URLSearchParams({ limit: String(limit), offset: String(offset) }).toString();
    const url = query ? `/users?${query}` : '/users';
    const response = await this.#request.get(url);

    const { users = [], total = 0 } = response || {};
    return [users.map(toDomain), total];
  }

  async assignRoles(userId, roles) {
    const response = await this.#request.put(
      `/users/${userId}/roles`,
      toAssignRolesDTO(roles),
    );

    return toDomain(response);
  }

  async getProfile() {
    const response = await this.#request.get('/profile');
    return toDomain(response);
  }
}

export default UsersRepository;
