# Access Control Architecture and Models

## Access Control Architecture
This system follows the **Policy-Based Access Control (PBAC)** model and consists of the following key components:

### Policy Enforcement Point (PEP)
- Receives incoming requests from users or applications and forwards them to the **Policy Decision Point (PDP)**.
- Enforces the decision (**allow** or **deny**) returned by the PDP.

### Policy Decision Point (PDP)
- Evaluates access requests based on predefined policies stored in the PRP and contextual information retrieved from the PIP.
- Determines whether a request should be granted or denied based on access control rules.

### Policy Retrieval Point (PRP)
- Stores access control policies that the PDP retrieves during evaluations.
- Acts as the **backend database** for access policies.

### Policy Information Point (PIP)
- Provides **external or dynamic data** required for policy evaluation.
- Examples include user attributes (e.g., department, security clearance), device information, or environmental conditions (e.g., time of access, location).

## Access Control Model

This system implements Role-Based Access Control (RBAC):

### Role-Based Access Control (RBAC)
RBAC assigns permissions based on predefined user roles. Users inherit permissions through their role assignments.

##### Example Use Case
Alice is a **Project Manager** at a company. In this system:
- The **Project Manager** role has permissions to **view**, **edit**, and **approve** project documents.
- Alice is assigned the **Project Manager** role.
- Because of this, Alice can edit and approve project-related documents.

##### How It Works for Alice
1. Alice logs into the system and tries to **edit** a project document.
2. The system checks Alice's **role** and sees that she is a **Project Manager**.
3. Since the **Project Manager** role includes the **edit** permission, Alice is allowed to proceed.
4. If Alice had been an **Employee** (who only has **view** permission), the system would have denied her request.
