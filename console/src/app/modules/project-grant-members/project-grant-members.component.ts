import { SelectionModel } from '@angular/cdk/collections';
import { AfterViewInit, Component, Input, OnInit, ViewChild } from '@angular/core';
import { MatDialog } from '@angular/material/dialog';
import { MatPaginator } from '@angular/material/paginator';
import { MatTable } from '@angular/material/table';
import { tap } from 'rxjs/operators';
import { ProjectMember, ProjectType } from 'src/app/proto/generated/management_pb';
import { ProjectService } from 'src/app/services/project.service';
import { ToastService } from 'src/app/services/toast.service';

import {
    ProjectGrantMembersCreateDialogComponent,
    ProjectGrantMembersCreateDialogExportType,
} from './project-grant-members-create-dialog/project-grant-members-create-dialog.component';
import { ProjectGrantMembersDataSource } from './project-grant-members-datasource';

@Component({
    selector: 'app-project-grant-members',
    templateUrl: './project-grant-members.component.html',
    styleUrls: ['./project-grant-members.component.scss'],
})
export class ProjectGrantMembersComponent implements AfterViewInit, OnInit {
    @Input() public projectId!: string;
    @Input() public grantId!: string;

    @Input() public type: ProjectType = ProjectType.PROJECTTYPE_GRANTED;

    @Input() public disabled: boolean = false;
    @ViewChild(MatPaginator) public paginator!: MatPaginator;
    @ViewChild(MatTable) public table!: MatTable<ProjectMember.AsObject>;
    public dataSource!: ProjectGrantMembersDataSource;
    public selection: SelectionModel<ProjectMember.AsObject> = new SelectionModel<ProjectMember.AsObject>(true, []);

    /** Columns displayed in the table. Columns IDs can be added, removed, or reordered. */
    public displayedColumns: string[] = ['select', 'firstname', 'lastname', 'username', 'email', 'roles'];

    public ProjectType: any = ProjectType;

    constructor(
        private projectService: ProjectService,
        private dialog: MatDialog,
        private toast: ToastService,
    ) { }

    public ngOnInit(): void {
        this.dataSource = new ProjectGrantMembersDataSource(this.projectService);
        console.log(this.projectId);
        this.dataSource.loadMembers(this.projectId, this.grantId, 0, 25, 'asc');
    }

    public ngAfterViewInit(): void {
        this.paginator.page
            .pipe(
                tap(() => this.loadMembersPage()),
            )
            .subscribe();
    }

    private loadMembersPage(): void {
        this.dataSource.loadMembers(
            this.projectId,
            this.grantId,
            this.paginator.pageIndex,
            this.paginator.pageSize,
        );
    }
    public removeProjectMemberSelection(): void {
        Promise.all(this.selection.selected.map(member => {
            return this.projectService.RemoveProjectGrantMember(this.projectId, this.grantId, member.userId).then(() => {
                this.toast.showInfo('Removed successfully');
            }).catch(error => {
                this.toast.showError(error.message);
            });
        }));
    }

    public removeMember(member: ProjectMember.AsObject): void {
        this.projectService.RemoveProjectGrantMember(this.projectId, this.grantId, member.userId).then(() => {
            this.toast.showInfo('Member removed successfully');
        }).catch(error => {
            this.toast.showError(error.message);
        });
    }

    public isAllSelected(): boolean {
        const numSelected = this.selection.selected.length;
        const numRows = this.dataSource.membersSubject.value.length;
        return numSelected === numRows;
    }

    public masterToggle(): void {
        this.isAllSelected() ?
            this.selection.clear() :
            this.dataSource.membersSubject.value.forEach(row => this.selection.select(row));
    }

    public async openAddMember(): Promise<any> {
        const keysList = (await this.projectService.GetProjectGrantMemberRoles()).toObject();
        console.log(keysList);

        const dialogRef = this.dialog.open(ProjectGrantMembersCreateDialogComponent, {
            data: {
                roleKeysList: keysList.rolesList,
            },
            width: '400px',
        });

        dialogRef.afterClosed().subscribe((dataToAdd: ProjectGrantMembersCreateDialogExportType) => {
            console.log(dataToAdd);
            if (dataToAdd) {
                dataToAdd.userIds.forEach((userid: string) => {
                    this.projectService.AddProjectGrantMember(
                        this.projectId,
                        this.grantId,
                        userid,
                        dataToAdd.rolesKeyList,
                    ).then(() => {
                        this.toast.showInfo('Project Grant Member successfully added!');
                    }).catch(error => {
                        this.toast.showError(error.message);
                    });
                });

            }
        });
    }
}